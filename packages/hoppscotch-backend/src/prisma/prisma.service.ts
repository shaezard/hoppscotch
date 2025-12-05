import { Injectable, OnModuleInit, OnModuleDestroy } from '@nestjs/common';
import { PrismaClient, Prisma } from 'src/generated/prisma/client';
import { PrismaPg } from '@prisma/adapter-pg';
import pg from 'pg';

@Injectable()
export class PrismaService
  extends PrismaClient
  implements OnModuleInit, OnModuleDestroy
{
  private pool: pg.Pool;

  constructor() {
    const pool = new pg.Pool({
      connectionString: process.env.DATABASE_URL,
    });
    const adapter = new PrismaPg(pool);

    super({
      adapter,
      transactionOptions: {
        maxWait: 5000, // 5 seconds
        timeout: 10000, // 10 seconds
      },
    });

    this.pool = pool;
  }
  async onModuleInit() {
    await this.$connect();
  }

  async onModuleDestroy() {
    await this.$disconnect();
    await this.pool.end();
  }

  /**
   * Locks rows in TeamCollection for a specific teamId and parentID.
   */
  async lockTeamCollectionByTeamAndParent(
    tx: Prisma.TransactionClient,
    teamId: string,
    parentID: string | null,
  ) {
    const lockQuery = parentID
      ? Prisma.sql`SELECT "orderIndex" FROM "TeamCollection" WHERE "teamID" = ${teamId} AND "parentID" = ${parentID} FOR UPDATE`
      : Prisma.sql`SELECT "orderIndex" FROM "TeamCollection" WHERE "teamID" = ${teamId} AND "parentID" IS NULL FOR UPDATE`;
    return tx.$executeRaw(lockQuery);
  }

  /**
   * Locks rows in TeamRequest for specific teamID and collectionIDs.
   */
  async lockTeamRequestByCollections(
    tx: Prisma.TransactionClient,
    teamID: string,
    collectionIDs: string[],
  ) {
    const lockQuery = Prisma.sql`SELECT "orderIndex" FROM "TeamRequest" WHERE "teamID" = ${teamID} AND "collectionID" IN (${Prisma.join(collectionIDs)}) FOR UPDATE`;
    return tx.$executeRaw(lockQuery);
  }

  /**
   * Locks rows in UserCollection for a specific userUid and parentID.
   */
  async lockUserCollectionByParent(
    tx: Prisma.TransactionClient,
    userUid: string,
    parentID: string | null,
  ) {
    const lockQuery = parentID
      ? Prisma.sql`SELECT "orderIndex" FROM "UserCollection" WHERE "userUid" = ${userUid} AND "parentID" = ${parentID} FOR UPDATE`
      : Prisma.sql`SELECT "orderIndex" FROM "UserCollection" WHERE "userUid" = ${userUid} AND "parentID" IS NULL FOR UPDATE`;
    return tx.$executeRaw(lockQuery);
  }

  /**
   * Locks rows in UserRequest for specific userUid and collectionIDs.
   */
  async lockUserRequestByCollections(
    tx: Prisma.TransactionClient,
    userUid: string,
    collectionIDs: string[] = [],
  ) {
    const lockQuery = Prisma.sql`SELECT "orderIndex" FROM "UserRequest" WHERE "userUid" = ${userUid} AND "collectionID" IN (${Prisma.join(collectionIDs)}) FOR UPDATE`;
    return tx.$executeRaw(lockQuery);
  }
}
